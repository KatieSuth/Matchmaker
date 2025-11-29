'use client'
import React, { useEffect, useState } from 'react'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faBars, faCaretDown } from '@fortawesome/free-solid-svg-icons'
import { usePathname } from 'next/navigation'
import Link from "next/link";
import styles from './Header.module.css';

const Header = ({discordName, discordId, image, isAdmin}) => {
    const [showNav, setShowNav] = useState(false)

    const hamburgerClick = (e) => {
        e.preventDefault()
        setShowNav(!showNav)
    }

    var eventsAdminActive = "",
        usersAdminActive = "",
        accountActive = "",
        eventsActive = "",
        showImage = false,
        imageUrl = ""


    if (usePathname().indexOf('/admin_events') >= 0) {
        eventsAdminActive = styles.active
    } else if (usePathname().indexOf('/admin_users') >= 0) {
        usersAdminActive = styles.active
    } else if (usePathname().indexOf('/my_account') >= 0) {
        accountActive = styles.active
    } else if (usePathname().indexOf('/events') >= 0) {
        eventsActive = styles.active
    }

    if (discordId.toString().length > 0 && image.length > 0) {
        showImage = true
        imageUrl = `https://cdn.discordapp.com/avatars/${discordId}/${image}`
    }

    return (
        <header>
            <div className={showNav ? `${styles.topnav} ${styles.responsive}` : styles.topnav}>
                <Link href="/events" className={styles.logo}><img src="/GalorantsBanner.jpg" alt="Galorants Banner" /></Link>
                <Link href="/events" className={eventsActive}>Events</Link>
                {isAdmin && (
                    <div className={styles.dropdown}>
                        <button className={styles.dropbtn}>Admin
                            <FontAwesomeIcon icon={faCaretDown} />
                        </button>
                        <div className={styles.dropdowncontent}>
                            <Link href="/admin_events" className={`${styles.dropdownLink} ${eventsAdminActive}`}>Events Admin</Link>
                            <Link href="/admin_users" className={`${styles.dropdownLink} ${usersAdminActive}`}>Users Admin</Link>
                        </div>
                    </div>
                )}
                <div className={`${styles.dropdown} ${styles.myaccount}`}>
                    <button className={styles.dropbtn}>
                        {showImage && (
                            <img className={styles.userimage} src={imageUrl} alt="User Profile Image" />
                        )}
                        <div className={styles.username}>
                            {discordName}
                            <FontAwesomeIcon icon={faCaretDown} />
                        </div>
                    </button>
                    <div className={`${styles.dropdowncontent} ${styles.myaccountcontent}`}>
                        <Link href="/my_account" className={`${styles.dropdownLink} ${accountActive}`}>My Account</Link>
                        <Link href="/logout" className={styles.dropdownLink}>Logout</Link>
                    </div>
                </div>
                <a href="true" className={styles.icon} onClick={hamburgerClick}>
                    <FontAwesomeIcon icon={faBars} />
                </a>
            </div>
            {/*TODO: figure out how to do this better with radix ui theme*/}
            <hr style={{color:"#262626"}} />
            <div className="mt-9"></div>
        </header>
    )
}

export default Header
